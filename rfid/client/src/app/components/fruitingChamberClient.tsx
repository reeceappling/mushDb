'use client'

import React, {useState} from "react";
import NotesArea, {NotesAreaOld, IsValidNote, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumbersOnlyFromText} from "@/app/components/formSubcomponents/date";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {AddToTransfers, InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    DisplayInput,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    MainCollectionInputOrRead,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    SendMultipartRequest,
    setFormData,
    setFormImages
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensInlineDisplay, GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea,
    SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SaleArea} from "@/app/components/saleClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {NewFruitForm} from "@/app/components/fruitClient";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {FruitData} from "@/app/components/fruitServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubstrateBatchArea, SubstrateBatchSelector} from "@/app/components/substrateBatchClient";
import {SelectorFor} from "@/app/components/selector";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {OvcForNewFruit} from "@/app/components/bagClient";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";

export function AssertFruitingChamber(input: any): asserts input is FruitingChamberData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
        ['cupsGrain', 'number'],
        ['mixedSubstratePerGrain', 'number'],
        ['casingPerGrain', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('FruitingChamber assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['substrateBatch', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('FruitingChamber assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        // ['entryType', (inp: any) => {
        //     return (typeof inp === 'string' && inp === "fruitingChamber")
        // }],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('FruitingChamber assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Fruiting chamber assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['flushes', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('FruitingChamber assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function FruitingChamberDisplay( // TODO: REDO WHOLE SECTION!
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertFruitingChamber(data)
        const [initial, setInitial] = useState(data)

        const existingNotes: Note[] = initial.notes || []
        const initNotes: Data<Note>[] = existingNotes.map((n) => {
            return {data: n, disabled: false}
        })
        // Basic objects
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [sale, setSale] = useState(initial.sale)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: initNotes, new: []})
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [writeTo, setWriteTo] = useState<string | undefined>()

        // Image-containing
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [flushes, setFlushes] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.flushes))

        // Helper states
        const [transfersOut, setTransfersOut] = useState(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: FruitingChamberData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setDisposed(updated.disposed)
            setSale(updated.sale)
            setNotes({existing: dataFor(updated.notes), new: []})
            setAcl(updated.acl)
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setFlushes(InitialPicsEntries(updated.flushes))
            setTransfersOut(updated.transfersOut || [])
        }
        const fruitingChamberSubmit = () => {
            let body = new FormData()
            let dataObj: any = {
                knownFruitable: knownFruitable,
                disposed: disposed,
                sale: sale,
                notes: notes,
                writeTagTo: writeTo,
                acl: MarshalAcl(acl),
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
                // Flushes
                let flushesInfo = resolvePicsFormData(flushes)
                let newFlushes = flushesInfo.images
                dataObj.flushes = flushesInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
                setFormImages(body, "newFlush", newFlushes)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/fruitingChamber/" + initial._id, cookies, body)
                .then(HandleJsonResponse).then((newEntry) => {
                try {
                    AssertFruitingChamber(newEntry)
                    updateInitial(newEntry)
                } catch (er) {
                    setErr("failed to decode response: " + JSON.stringify(er))
                }
            }).catch((er) => {
                setErr(JSON.stringify(er))
            });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            OvcForNewFruit(data._id, "fruitingChamber", cookies),
            // TODO: xfers? OvcForXfers(data._id, "fruit", ["plate","slant","jar","stasisTube"], "Clone/Transfer Fruit"), // TODO: ensure list correct// TODO: OVC for clone to plate (transfer)
            // TODO: spore swab directly from box? (should also create a fruit in the interim)
            // TODO: spore print directly from box? (should also create a fruit in the interim)
        ]
        return (
            <DisplayFormWrapper entryType={"fruitingChamber"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Fruiting Chamber"} entryType={"fruitingChamber"}/>
                <MostRecentImageDisplay data={initial.mostRecentImage}
                                        headerLevel={headerLevel}/>{/* Most recent image! */}
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={initial.disposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SubstrateRecipeArea id={initial.recipe} headerLevel={headerLevel} readonly={true}/>
                        <SubstrateBatchArea id={initial.substrateBatch} readonly={true}/>
                        <div>{"Cups grain: " + data.cupsGrain}</div>
                        <div>{"Substrate mixed with grain: " + (data.mixedSubstratePerGrain * data.cupsGrain)}</div>
                        <div>{"Casing: " + (data.casingPerGrain * data.cupsGrain)}</div>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore}
                                            gensSinceFruitOrSpore={initial.genFruitOrSpore}
                                            headerLevel={headerLevel}/>{/* Gens since spore and spore/fruit*/}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                        <InnocDisplay innoc={initial.innoc} openInNewTab={false}/>{/* Innoc */}
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>{/* Parent */}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <KnownFruitableArea initial={knownFruitable} readonly={readonly} headerLevel={headerLevel}
                                            doSelect={setKnownFruitable}/>{/* KnownFruitable */}
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                                  canCreateSale={true}/>{/* Sale */}
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"fruitingChamber"}
                                     transfersOut={initial.transfersOut} allowNewTransferCreation={false}
                                     cookies={cookies}/>

                <PicsDisplay pix={pics} readonly={readonly} updateParent={setPics}/>{/* Pics */}
                {/* Flushes */}<PicsDisplay pix={flushes} readonly={readonly}
                                            updateParent={setFlushes} addButtonText={"Create New Flush"}/>
                <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams}
                                readonly={readonly} headerLevel={headerLevel}/>{/* Contamination */}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {/*<EntryPermsArea originalPerms={initial.perms} setEntryPerms={setPerms}/>*/}
                {readonly ? null :
                    <button className={"bottomButton greenButton"} onClick={(e)=>{
                        e.stopPropagation();
                        fruitingChamberSubmit()
                    }}>{"Update"}</button>}
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Fruiting Chamber data format incorrect: " + err}</div>
    }
}

// TODO: cups/pints/quarts selector for finding number of cups volume
export function VolumeSelector({initialVal, initialUnit, updateNumberOfCups}: {
    initialVal?: number,
    initialUnit?: string,
    updateNumberOfCups: (n: number) => void
}) {
    // TODO: INITIAL NUMBERS AND UNITS
    const mulForUnit = (u: string) => {
        switch (u) {
            case "cups":
                return 1
            case "pints":
                return 2
            case "quarts":
                return 4
            default:
                return 0
        }
    }
    const [u, setUnit] = useState(initialUnit || "quarts")
    const [val, setVal] = useState<number>(initialVal || 1)
    const updateNumber = (s: string) => {
        let n = NumbersOnlyFromText(s)
        updateNumberOfCups(n * mulForUnit(u))
        setVal(n)
    }

    const changeUnit = (s: string) => {
        updateNumberOfCups(val * mulForUnit(s))
        setUnit(s)
    }

    updateNumberOfCups(val * mulForUnit(u))
    return <div>
        <input className={"volNumInput"} type="text" value={val} onChange={(e) => {
            updateNumber(e.currentTarget.value)
        }}/>
        {/* Volume unit selector */}
        <SelectorFor options={["cups", "pints", "quarts"]} initial={u} updateParent={changeUnit} disabled={false}/>
    </div>
}

// TODO: MOVE THIS!
export function FloatInput({initial, onChange}: { initial?: number, onChange: (value: number) => void }) {
    const [val, setVal] = useState<number>(initial || 0)
    const updateNumber = (s: string) => {
        let n = NumbersOnlyFromText(s)
        setVal(n)
        onChange(n)
    }
    return <div>
        <input className={"volNumInput"} type="text" value={val} onChange={(e) => {
            updateNumber(e.currentTarget.value)
        }}/>
    </div>
}

export function NewFruitingChamberForm({handlers, substrateBatchIn, parent}: {
    handlers: NewEntryInput<FruitingChamberData>,
    substrateBatchIn?: SubstrateBatchData,
    parent?: string
}) {
    const [subBatch, setSubBatch] = useState<SubstrateBatchData | undefined>(substrateBatchIn)
    const [parentId, setParentId] = useState<string | undefined>() // TODO: do we even want this in the initial one? Keep it optional
    const [volumeGrainCups, setVolumeGrainCups] = useState<number>(0)
    const [mixedSubCups, setMixedSubCups] = useState<number>(0)
    const [casingCups, setCasingCups] = useState<number>(0)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const newFruitingChamberSubmit = () => {
        if (!subBatch) {
            setErr("substrate batch must be selected")
            return
        }
        if (volumeGrainCups <= 0) {
            setErr("casing must be a positive value")
            return
        }
        if (casingCups < 0) {
            setErr("casing must be a positive or zero value")
            return
        }
        if (mixedSubCups <= 0) {
            setErr("mixed substrate must be a positive number")
            return
        }
        let body: any = {
            substrateBatch: subBatch._id,
            parentId: parentId,
            grainCups: volumeGrainCups,
            mixedSubCups: mixedSubCups,
            casingCups: casingCups,
            notes: notes
        }
        writeTagTo && (body.writeTagTo = writeTagTo)
        fetch(BaseExternalUrl + "/create/fruitingChamber", {
            method: 'Post',
            body: JSON.stringify(body),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                try {
                    AssertFruitingChamber(newEntry)
                    handlers.onCreate && handlers.onCreate(newEntry)
                } catch (er) {
                    throw new Error("failed to decode response: " + JSON.stringify(er))
                }
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    const batchArea = () => {
        if (substrateBatchIn) {
            return <SubstrateBatchArea readonly={true} id={substrateBatchIn._id}/>
        }
        return <SubstrateBatchSelector doSelect={(sb => {
            setSubBatch(sb)
        })} allowCreation={false/* TODO: ok?*/} creatorInPage={false/*TODO:????*/}/>
    }
    const parentArea = () => {
        if (!parent) {
            {/* TODO: input b58 id or read from rfid*/
            }
            return <MainCollectionInputOrRead onIdSelected={setParentId}/>
        }
        return <div>{parentId || "unknown"}</div> // TODO: FIX
    }
    return (
        <NewEntryFormWrapper entryType={"fruitingChamber"}>
            <ErrorDisplay err={err}/>
            {batchArea()}
            {/* TODO: grain volume from parent????*/}
            {parentArea()}
            {"Volume of grain mixed: "}<VolumeSelector initialVal={0} initialUnit={"quarts"}
                                                       updateNumberOfCups={setVolumeGrainCups}/>
            {"Volume of substrate mixed: "}<VolumeSelector initialVal={0} initialUnit={"quarts"}
                                                           updateNumberOfCups={setMixedSubCups}/>
            {"Volume casing: "}<VolumeSelector initialVal={0} initialUnit={"quarts"}
                                               updateNumberOfCups={setCasingCups}/>
            <NotesAreaOld readonly={false} current={{ // TODO: notesFormArea?
                existing: [], new: notes.map(v => {
                    return {data: v, disabled: false}
                })
            }} updateParent={(nots) => {
                setNotes(nots.new.map((n) => {
                    return n.data
                }))
            }}/>
            <ReaderWriterSelector onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" onClick={newFruitingChamberSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function FruitingChamberImportDisplay({headerLevel, cookies}: ImportDisplayInput) {
    const [recipe, setRecipe] = useState<SubstrateRecipeData | undefined>()
    const [creationDate, setCreationDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [grainCups, setGrainCups] = useState<number>(4) // TODO: set
    const [mixedSubRatio, setMixedSubRatio] = useState<number>(4) // TODO: set
    const [casingRatio, setCasingRatio] = useState<number>(2) // TODO: set
    // Non-required
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    const submitImportFruitingChamber = () => {
        const reqd = new Map<string, any>([
            ['recipe', recipe],
            ['creationDate', creationDate],
            ['species', species]
        ])
        for (let [key, val] of reqd) {
            if (val === undefined) {
                setErr(key + " must be defined!");
                return
            }
        }

        let bodyObj: any = {
            recipe: recipe,
            creationDate: creationDate,
            grainCups: grainCups,
            substrateRatio: mixedSubRatio,
            casingRatio: casingRatio,
            species: species?._id,
            //perms: perms,
        }
        subspecies && (bodyObj.subspecies = subspecies._id)
        generation && (bodyObj.generation = generation)
        knownFruitable && (bodyObj.knownFruitable = knownFruitable)
        writeTagTo && (bodyObj.writeTagTo = writeTagTo)
        let formData = new FormData()

        imageFile && formData.set("img", imageFile, "img")
        setFormData(formData, bodyObj)

        fetch(BaseExternalUrl + "/db/import/fruitingChamber", {
            method: 'Post',
            body: formData,
            headers: {
                credentials: 'include',
                'Cookie': cookies,
                'Content-type': "multipart/form-data"
            },
        }).then(HandleTxtResponse).then((newId) => {
            redirect(BaseExternalUrl + "/view/fruitingChamber/" + newId)
        }).catch((err) => {
            setErr(JSON.stringify(err))
        });
    }
    return <ImportEntryFormWrapper entryType={"fruitingChamber"}>
        <ErrorDisplay headerLevel={headerLevel} err={err}/>
        {/* Required Fields */}
        <SubstrateRecipeSelector doSelect={setRecipe} headerLevel={headerLevel}/>
        <DateArea pre={"Created on: "} readonly={false} when={Date.now()} updateParent={setCreationDate}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
        {"Grain volume: "}<VolumeSelector initialVal={1} initialUnit={"quarts"} updateNumberOfCups={setGrainCups}/>
        {"Mixed substrate ratio"}<FloatInput initial={mixedSubRatio} onChange={setMixedSubRatio}/>
        {"Casing ratio"}<FloatInput initial={casingRatio} onChange={setCasingRatio}/>
        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                    headerLevel={headerLevel/*cookies={cookies}*/}/>
        <GenerationArea readonly={false} updateParent={setGeneration} headerLevel={headerLevel}/>
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImageFile}/>
        {/*<EntryPermsArea setEntryPerms={setPerms}/>*/}
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        {/* SUBMIT AREA */}
        <input type="submit" value="Submit" onClick={submitImportFruitingChamber} onSubmit={(e) => {
            e.preventDefault();
        }}/>
    </ImportEntryFormWrapper>
}

export function FruitingChamberInline({
                                          data,
                                          expandByDefault,
                                          onClick,
                                          showMainPageButton,
                                          idIsLink
                                      }: InlineProps<FruitingChamberData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const base58Id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={base58Id} txt={"Fruiting Chamber"} entryType={"fruitingChamber"}
                allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            {/* TODO: ADD substrateRatios (mixed and casing)*/}
            <MostRecentImageDisplay data={data.mostRecentImage}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} initialSub={data.subspecies}
                            currentSpecies={data.species}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore}/>
            <SubstrateRecipeArea id={data.recipe} readonly={true}/>
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <DisposedDisplay disposed={data.disposed} readonly={true}/>{/* TODO: SOLD AND DISPOSED DISPLAY INLINE */}
            <SaleArea sale={data.sale} readonly={true} canCreateSale={false}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            {/* TODO: PARENT ID AREA */}
            <SubstrateBatchArea readonly={true}/>
            {/*TODO: <ProjectsArea projects={data.perms?.projectPerms.ids} headerLevel={headerLevel} offset={-1}*/}
            {/*              allowCreate={false} readonly={true} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function FruitingChamberListDisplay({data, onClick}: SingleListProps<FruitingChamberData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <FruitingChamberInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }