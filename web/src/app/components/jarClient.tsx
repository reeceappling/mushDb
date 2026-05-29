'use client'

import {JarData} from "@/app/components/jarServer";
import {
    DisplayFormWrapper,
    DisplayInput, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    importApiUrlFor,
    ImportDisplayInput, ImportEntryFormWrapper,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData, SelectorWrapper,
    SendMultipartRequest,
    setFormData,
    setFormImages, viewUrlFor
} from "@/app/components/common";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import React, {JSX, useState} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import ID from "@/app/components/formSubcomponents/id";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {JarRecipeArea, JarRecipeSelector} from "@/app/components/jarRecipeClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {PcRunArea} from "@/app/components/pcRunClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {BaseExternalUrl} from "@/app/components/Constants";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    InitialNotesState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {SaleArea} from "@/app/components/saleClient";
import {
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {redirect} from "next/navigation";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {JarSizeSelector} from "@/app/components/formSubcomponents/utils/volumeSelector";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {GrainBatchData, GrainBatchSelectorCloseable} from "@/app/components/grainBatchServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";

export function AssertJar(input: any): asserts input is JarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['sizeCups', 'number'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
        ['sizeCups', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Jar assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['burstGrains', 'number'],
        ['wetness', 'number'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['grainBatch', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        // ['entryType', (inp: any) => {
        //     return (typeof inp === 'string' && inp === "jar")
        // }],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Jar assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Jar assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function JarImportDisplay({headerLevel, cookies}: ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [recipe, setRecipe] = useState<JarRecipeData | undefined>()
    const [sizeCups, setSizeCups] = useState<number>(4)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const importEntry = () => {
        let formData = new FormData()
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        if (recipe === undefined) {
            setErr("Recipe must be set!")
            return
        }
        let dataObj: any = {
            created: created,
            sizeCups: sizeCups,
            recipe: recipe._id,
            species: species._id,
            //perms: perms,
        }
        subspecies && (dataObj.subspecies = subspecies._id)
        knownFruitable && (dataObj.knownFruitable = knownFruitable)
        generation && (dataObj.generation = generation)
        if (imageFile !== undefined) {
            formData.set("img", imageFile, "img")
        }
        writeTagTo && (dataObj.writeTagTo = writeTagTo)


        SendMultipartRequest(importApiUrlFor("jar"), cookies, formData)
            .then(HandleJsonResponse)
            .then((newItem) => {
                AssertJar(newItem)
                redirect(viewUrlFor("jar",newItem._id))
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"jar"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <div className={"inlineChildren"}>
            <div className={"mr-2"}>{"Size: "}</div>
            <JarSizeSelector onChange={(s: string) => {
                if (s === "pint") {
                    setSizeCups(2)
                } else if (s === "quart") {
                    setSizeCups(4)
                } else {
                    setErr("invalid size cups")
                }
            }}/>
        </div>
        <SelectorWrapper current={recipe} title={"Jar Recipe"} nameFunc={(v: JarRecipeData) => v._id}>
            <JarRecipeSelector doSelect={setRecipe} allowCreate={true}/>
        </SelectorWrapper>
        <ExistingSpeciesSelector doSelect={setSpecies/*cookies={cookies}*/}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies/*cookies={cookies}*/}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={importEntry}>{"Import"}</button>
    </ImportEntryFormWrapper>
}

function sizeFromNum(cups: number) {
    switch (cups) {
        case 1:
            return "cup"
        case 2:
            return "pint"
        case 4:
            return "quart"
        default:
            return "unhandled number of cups (" + cups + ")"
    }
}

function cupsPer(unit: string) {
    switch (unit) {
        case "cup":
            return 1
        case "pint":
            return 2
        case "quart":
            return 4
        case "gallon":
            return 16
        default:
            return -1
    }
}

export default function JarDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {


    try {
        AssertJar(data)
        const [initial, setInitial] = useState(data)

        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        // TODO: wetness (but can only be set once)
        // TODO: burst grains (but can only be set once)

        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: JarData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setAcl(updated.acl)
            // TODO: wetness (but can only be set once)
            // TODO: burst grains (but can only be set once)
            setTransfersOut(updated.transfersOut || [])
        }
        const submit = () => {
            let body = new FormData()
            let dataObj: any = {
                knownFruitable: knownFruitable,
                disposed: disposed,
                sale: sale,
                writeTagTo: writeTagTo,
                acl: MarshalAcl(acl),
                notes: notes,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data", JSON.stringify(dataObj))
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/jar/" + initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertJar(newEntry)
                    updateInitial(newEntry)
                })
                .catch((er) => {
                    setErr("failed to decode response: " + JSON.stringify(er))
                });
        }
        const jarSizeArea = () => {
            return <div>
                {"Size: " + sizeFromNum(data.sizeCups)}
            </div>
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
        ]
        return <DisplayFormWrapper entryType={"jar"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={data._id} txt={"Grain Jar"} entryType={"jar"}/>
            <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                disposed={disposed} setDisposedOnParent={setDisposed}
                                                readonly={readonly}/>
                    {jarSizeArea()}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <JarRecipeArea headerLevel={headerLevel} recipeId={initial.recipe}/>
                    <PcRunArea binaryId={initial.pcRun} headerLevel={headerLevel}/>

                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    <InnocDisplay innoc={initial.innoc}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}
                                        headerLevel={headerLevel}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                              canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}
                                     headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>

            <TransfersOutDisplay thisId={initial._id} thisEntryType={"jar"} transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly} cookies={cookies}/>
            <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>
            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>

            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null :
                <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>}
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Grain Jar data format incorrect: " + err}</div>
    }
}

// NewJarForm is used from the recipe page. PcRun CAN be created from here?
// TODO: NewJarForm is used from the recipe page. PcRun CAN be created from here?
export function NewJarForm({handlers, recipeIn, pcRunIn, grainBatchIn}: {
    handlers: NewEntryInput<JarData>,
    recipeIn?: string,
    pcRunIn?: PcRunData,
    grainBatchIn?: GrainBatchData
}) {
    const [creationDate, setCreationDate] = useState(Date.now()) // TODO: use?
    const [grainBatch, setGrainBatch] = useState(grainBatchIn) // TODO: use?
    const [recipe, setRecipe] = useState(recipeIn) // TODO: use?
    const [sizeCups, setSizeCups] = useState<number>(4) // TODO: change initial state?
    const [pcRun, setPcRun] = useState(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])

    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>()
    const createJar = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!recipe || !pcRun) {
            setErr("recipe and pc run must both exist!")
            return
        }
        if (sizeCups < 1) {
            setErr("must select a valid jar volume")
            return
        }
        fetch(BaseExternalUrl + "/db/create/jar", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify({
                //creationDate: creationDate, // TODO: GET FROM PC RUN! handle in go
                sizeCups: sizeCups,
                recipe: recipe, // TODO: handle properly in go
                batch: grainBatch?._id, // TODO: handle properly in go
                pcRun: pcRun._id,
                notes: notes || [],
                writeTagTo: writeTagTo,
            })
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertJar(newEntry)
                handlers.onCreate && handlers.onCreate(newEntry)
            })
            .catch((error) => {
                setErr("failed to unmarshal create jar response: " + JSON.stringify(error))
            });
    }
    const hasGrainBatchOrRecipe = grainBatchIn !== undefined || recipeIn !== undefined
    return <NewEntryFormWrapper entryType={"jar"}>
        <ErrorDisplay err={err}/>
        {/* TODO: REMOVE? */}<DateArea pre={"Creation date: "} when={Date.now()} readonly={false}
                                       updateParent={setCreationDate}/>
        {/* TODO: BATCH!!!!*/}
        {/* TODO: PICK BETWEEN NEXT 2! */}
        {hasGrainBatchOrRecipe && <GrainBatchSelectorCloseable doSelect={setGrainBatch}
                                                               allowCreation={handlers.isTopLevel} creatorInPage={handlers.isTopLevel}/>}
        {/*{hasGrainBatchOrRecipe && <GrainBatchSelector doSelect={setGrainBatch}*/}
        {/*                                              allowCreate={handlers.isTopLevel}/>}*/}
        {hasGrainBatchOrRecipe && // TODO: validate ok
            <JarRecipeSelector allowCreate={handlers.isTopLevel} doSelect={(rec?: JarRecipeData) => {
                setRecipe(rec?._id)
            }}/>} {/* TODO: CreatorInPage reference from non-isTopLevel. CLOSEABLE???*/}
        <JarSizeSelector onChange={(unit: string) => {
            setSizeCups(cupsPer(unit))
        }}/>
        {pcRunIn !== undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                                          creatorInPage={handlers.isTopLevel}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createJar}>{"Submit new Jar"}</button>
    </NewEntryFormWrapper>
}

export function JarListPageTable({data, onClick, withLink}: ListPageItems<JarData>) {
    let cols: ListTableColumn<JarData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: JarData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "jar", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function JarSelectorTable({data, onClick}: ListPageItems<JarData>) {
    return <JarListPageTable data={data} onClick={onClick}/>
}

export function JarSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: JarData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: JarData[]): JSX.Element => {
        return <JarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"jar"} entryTypes={"jars"} doSelect={doSelect} asserter={AssertJar}
                                   table={table}>
        {allowCreate && <NewJarForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
