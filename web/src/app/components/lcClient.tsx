'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcData} from "@/app/components/lcServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    ConfirmedCleanArea,
    ConfirmedCleanSelector,
    createApiUrlFor,
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateMultipartRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    HandleTxtResponse,
    importApiUrlFor,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    resolveContamsFormData,
    resolvePicsFormData,
    setFormData,
    setFormImages,
    updateApiUrlFor,
    viewUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {BaseExternalUrl} from "@/app/components/Constants";
import {LcRecipeData, LcRecipeSelectorCloseable} from "@/app/components/lcRecipeServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import ID from "@/app/components/formSubcomponents/id";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {NewLcSyringeForm} from "@/app/components/lcSyringeClient";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {LcRecipeArea} from "@/app/components/lcRecipeClient";
import {AssertJar} from "@/app/components/jarClient";
import {AssertJarRecipe} from "@/app/components/jarRecipeClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";

export function AssertLc(input: any): asserts input is LcData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Lc assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['confirmedClean', 'boolean'],
        ['knownFruitable', 'boolean'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Lc assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl],
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Lc assertion failure: optional key ' + key + ' was not valid');
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
            throw new Error('Lc assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function LcImportDisplay({headerLevel}: ImportDisplayInput) {
    const [recipe, setRecipe] = useState<LcRecipeData | undefined>(undefined)
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const ImportLc = () => {
        let formData = new FormData()
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        if (recipe === undefined) {
            setErr("Recipe must be set!")
            return
        }
        let bodyObj: any = {
            creationDate: created,
            species: species._id,
            recipe: recipe._id,
            // Optionals
            subspecies: subspecies?._id,
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
        }
        setFormData(formData, bodyObj)
        //formData.set("data", JSON.stringify(bodyObj))
        if (imageFile !== undefined) {
            formData.set("img", imageFile, "img")
        }

        MultipartImportRequest(formData, "lc", AssertLc, setErr, allCookies(cookies))
        // SendMultipartRequest(importApiUrlFor("lc"), cookies, formData) // TODO: remove cookies from call?
        //     .then(HandleJsonResponse)
        //     .then(newItem => {
        //         AssertLc(newItem)
        //         redirect(viewUrlFor("lc", newItem._id))
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"lc"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <LcRecipeSelectorCloseable doSelect={setRecipe} txt={"Recipe"} creatorInPage={false/* TODO: ok?*/} allowCreation={true} />
        {/*<SelectorWrapper current={recipe} title={"Recipe"} nameFunc={(v: LcRecipeData) => v._id}>*/}
        {/*    <LcRecipeSelector doSelect={setRecipe} allowCreate={true}/>/!* TODO: OPEN/CLOSE!*!/*/}
        {/*</SelectorWrapper>*/}
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                                                    headerLevel={headerLevel}/>
        <ConfirmedCleanSelector updateParent={setConfirmedClean}/>
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}
                              headerLevel={headerLevel}/>
        <button className={"greenButton buttonFullWidth"} onClick={ImportLc}>{"Import"}</button>
    </ImportEntryFormWrapper>
}

export default function LcDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertLc(data)
        const [initial, setInitial] = useState(data)

        const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(data.confirmedClean)
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: LcData) => {
            setInitial(updated)
            setConfirmedClean(updated.confirmedClean)
            setTransfersOut(updated.transfersOut || [])
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const lcSubmit = () => {
            let formData = new FormData()
            let bodyObj: any = {
                notes: notes,
                // Optionals
                confirmedClean: confirmedClean,
                knownFruitable: knownFruitable,
                disposed: disposed,
                writeTagTo: writeTagTo,
                acl: acl,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(images)
                let newImages = picsInfo.images
                bodyObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                bodyObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(formData, bodyObj)
                setFormImages(formData, "newPic", newImages)
                setFormImages(formData, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateMultipartRequest("lc",initial._id, formData, AssertLc, allCookies(cookies))
                .then(v=>{
                    updateInitial(new LcData(v))
                })
                .catch(e=>{
                    setErr(JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New Liquid Culture Syringe",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewLcSyringeForm txt={"Create New Liquid Culture Syringe"} parentLc={initial}
                                             onCreate={(lcs: LcSyringeData) => { // TODO: should swap to handler={{}} format rather than direct onCreate
                        onCreate([{
                            typeText: "Liquid Culture Syringe",
                            node: <CreatedLinkFor linkId={lcs._id} typ={"lcSyringe"}/>,// TODO: ENSURE lcs or lcSyringe is correct here
                        }], false)
                    }}/>
                },
            },
            // TODO: sale?
            WriteRfidOvcArea(initial._id),
        ]
        return <DisplayFormWrapper entryType={"lc"}>
            <ID txt={"Liquid Culture"} id={data._id} entryType={"lc"}/>
            {readonly ||
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>}{/* TODO: where to put?*/}
            <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    {/*3high,19wide fragmented*/}
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                disposed={disposed} setDisposedOnParent={setDisposed}
                                                readonly={readonly}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {/*1-2high (fragmented), extremely variable wide*/}
                    <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                    {/*1high (13-25)wide */}
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {/*3high 17-18 wide*/}
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {/*1high (19-25 wide)*/}
                    <ConfirmedCleanArea onSelect={setConfirmedClean} readonly={readonly}
                                        initial={initial.confirmedClean}/>
                    {/*1high (21-24 wide)*/}
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                </FlexedSinglesGroup>


                <FlexedSinglesGroup>
                    {/*1high (18+altId)*/}
                    <InnocDisplay innoc={initial.innoc}/>
                    {/*1high (26+altId)*/}
                    <LcRecipeArea lcRecipeId={initial.recipe}/>
                    {/*1high (10+(altId/7))*/}
                    <PcRunArea binaryId={initial.pcRun}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plate"}
                                 transfersOut={initial.transfersOut}
                                 allowNewTransferCreation={!readonly}/>
            <PicsDisplay pix={initial.pics || []} updateParent={setImages} readonly={readonly}
                         headerLevel={headerLevel}/>{/* Pics */}
            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>


            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly || <>
                <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
                <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    lcSubmit()
                }}>{"Update"}</button>
            </>}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Liquid Culture data format incorrect: " + err}</div>
    }
}

export function NewLcForm({handlers, lcRecipeIn, pcRunIn}: {
    handlers: NewEntryInput<LcData>,
    lcRecipeIn?: LcRecipeData,
    pcRunIn?: PcRunData
}) {
    const [creationDate, setCreationDate] = useState(Date.now())
    const [lcRecipe, setLcRecipe] = useState(lcRecipeIn)
    const [pcRun, setPcRun] = useState(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (lcRecipe === undefined) {
            setErr("A recipe must be selected")
            return
        }
        if (pcRun === undefined) {
            setErr("A PC Run must be selected")
            return
        }
        let body: any = {
            creationDate: creationDate,
            recipe: lcRecipe,
            pcRun: pcRun,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("lc", body, AssertLc, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }

    return <NewEntryFormWrapper entryType={"lc"}>
        <ErrorDisplay err={err}/>
        {lcRecipeIn !== undefined && <LcRecipeSelectorCloseable doSelect={setLcRecipe}
                                                       allowCreation={handlers.isTopLevel} creatorInPage={handlers.isTopLevel}/>} {/* TODO: isTopLevel? disallow ok? */}{/* TODO: closeable or no? */}
        {pcRunIn !== undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                                          creatorInPage={true}/>} {/* TODO: isTopLevel? disallow ok? */}
        <DateArea pre={"Creation date: "} when={Date.now()} readonly={false} updateParent={setCreationDate}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function LcListPageTable({data, onClick, withLink}: ListPageItems<LcData>) {
    let cols: ListTableColumn<LcData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Clean", v => v.confirmedClean ? (v.confirmedClean ? "clean" : "dirty") : "?"),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new LcData(v)}}/>
}

export function LcSelectorTable({data, onClick}: ListPageItems<LcData>) {
    let cols: ListTableColumn<LcData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Made", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Clean", v => v.confirmedClean ? (v.confirmedClean ? "clean" : "dirty") : "?"),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: LcData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new LcData(v)}}/>
}

export function LcSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: LcData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: LcData[]): JSX.Element => {
        return <LcSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"lc"} entryTypes={"lcs"} doSelect={doSelect} asserter={AssertLc}
                                   table={table}>
        {allowCreate && <NewLcForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
