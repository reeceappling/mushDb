'use client'

import {useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
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
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {
    ConfirmedCleanArea,
    ConfirmedCleanSelector,
    DisplayInput,
    DisposedContamArea,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    resolveContamsFormData,
    resolvePicsFormData,
    SendMultipartRequest,
    setFormData,
    setFormImages,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensFormDisplay,
    GensInlineDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea,
    SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState, InitialNotesState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    DisplayFormWrapper,
    ImportEntryFormWrapper,
    LcRecipeArea,
    LcRecipeSelector,
    NewEntryFormWrapper
} from "@/app/components/lcRecipeClient";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import ID from "@/app/components/formSubcomponents/id";
import {AddToTransfers, InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {OnViewCreatorsQuadColArea, PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, RecentPCRunSelector} from "@/app/components/pcRunServer";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {NewLcSyringeForm} from "@/app/components/lcSyringeClient";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";

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

export function LcImportDisplay({headerLevel, cookies}: ImportDisplayInput) {
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
            subspecies: subspecies?._id, // TODO: ok? // TODO: ensure correct capitalization
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

        SendMultipartRequest(BaseExternalUrl + "/db/import/lc", cookies, formData)
            .then(HandleTxtResponse)
            .then((newLcId) => {
                redirect(BaseExternalUrl + "/view/lc/" + newLcId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"lc"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <LcRecipeSelector doSelect={setRecipe} allowCreation={true} creatorInPage={true} headerLevel={headerLevel}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                    headerLevel={headerLevel/*cookies={cookies}*/}/>
        <ConfirmedCleanSelector selProps={{doSelect: setConfirmedClean, headerLevel: headerLevel}}/>
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <GenerationArea readonly={false} updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}
                              headerLevel={headerLevel}/>
        <button className={"greenButton"} onClick={ImportLc}></button>
    </ImportEntryFormWrapper>
}

export default function LcDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
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
        const lcSubmit = () => {
            let body = new FormData()
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
                setFormData(body, bodyObj)
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/lc/" + initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((updatedEntry) => {
                    AssertLc(updatedEntry)
                    updateInitial(updatedEntry)
                })
                .catch((err) => {
                    setErr(JSON.stringify(err))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: TEST HEAVILY
            {
                txt: "New Liquid Culture Syringe",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewLcSyringeForm txt={"Create New Liquid Culture Syringe"} parentLc={initial}
                                             cookies={cookies} onCreate={(lcs: LcSyringe) => { // TODO: should swap to handler={{}} format rather than direct onCreate
                        onCreate([{
                            typeText: "Liquid Culture Syringe",
                            node: <CreatedLinkFor linkId={lcs._id} typ={"lcSyringe"}/>,// TODO: ENSURE lcs or lcSyringe is correct here
                        }])
                    }}/>
                }
            },
            // TODO: can lc do anything else?
        ]
        return <DisplayFormWrapper entryType={"lc"}>
            <ID txt={"Liquid Culture"} id={data._id} entryType={"lc"}/>
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
            <TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plate"} transfersOut={initial.transfersOut}
                                 allowNewTransferCreation={!readonly}
                                 cookies={cookies}/>
            <PicsDisplay pix={images} updateParent={setImages} readonly={readonly}
                         headerLevel={headerLevel}/>{/* Pics */}
            <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>


            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly || <>
                <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
                <button className={"bottomButton"} onClick={lcSubmit}>{"Update"}</button>
            </>}
            {readonly ||
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>}{/* TODO: where to put?*/}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Liquid Culture data format incorrect: " + err}</div>
    }
}

export function SpeciesSubspeciesArea({species, subspecies}: {
    subspecies?: string,
    species?: string,
}) {
    const specArea = <div>
        {"Species: " + (species ? species : "none")}{/* TODO: LINK!?*/}
    </div>
    if (subspecies) {
        return <>
            {/* <div className={"specSubspecArea"}> */}
            {specArea}
            <div>
                {"Subspecies: " + subspecies}{/* TODO: LINK!*/}
            </div>
            {/*</div>*/}
        </>
    }
    return <>{/* <div className={"specSubspecArea"}> */}
        {specArea}
        {/*</div>*/}
    </>
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
        let body: any = {creationDate: creationDate, recipe: lcRecipe, pcRun: pcRun}
        notes && (body.notes = notes)
        writeTagTo && (body.writeTagTo = writeTagTo)
        fetch(BaseExternalUrl + "/create/lc", { // TODO: ensure correct
            method: "POST",
            headers: {
                credentials: 'include',
                // 'Cookie': cookies, // TODO: do we even need this here?
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertLc(newEntry)
                handlers.onCreate && handlers.onCreate(newEntry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }

    return <NewEntryFormWrapper entryType={"lc"}>
        <ErrorDisplay err={err}/>
        {lcRecipeIn !== undefined && <LcRecipeSelector doSelect={setLcRecipe} allowCreation={handlers.isTopLevel}
                                                       creatorInPage={true}/>} {/* TODO: isTopLevel? disallow ok? */}
        {pcRunIn !== undefined && <RecentPCRunSelector doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                                       creatorInPage={true}/>} {/* TODO: isTopLevel? disallow ok? */}
        <DateArea pre={"Creation date: "} when={Date.now()} readonly={false} updateParent={setCreationDate}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function LcInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<LcData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} entryType={"lc"} txt={"Liquid Culture Jar"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <InnocDisplay innoc={data.innoc} openInNewTab={true}/>
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} initialSub={data.species} currentSpecies={data.species}
            />
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <ConfirmedCleanArea initial={data.confirmedClean} readonly={true}/>
            <DisposedContamArea contams={data.contamination} disposed={data.disposed}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <LcRecipeArea lcRecipeId={data.recipe}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore} offset={-1}/>
            {/*TODO: <ProjectsArea allowCreate={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel} offset={-1} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function LcInline2({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<LcData>) { // TODO: REDO THIS!
//     const [expanded, setExpanded] = useState(expandByDefault)
//     const b58id = data._id
//     return <InlineEntry onClick={onClick}>
//         <InlineSubGroup >
//         <ID id={b58id} entryType={"lc"} txt={"Liquid Culture Jar"} allowOpenMainPage={showMainPageButton}
//             linkPage={idIsLink}/>
//         <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                expanded={expanded}/>
//         </InlineSubGroup>
//         <InlineSubGroup >
//         <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
//         <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         <DisposedContamArea contams={data.contamination} disposed={data.disposed}/>
//         <SpeciesArea readonly={true} initial={data.species}/>
//         <SubspeciesArea readonly={true} initialSub={data.species} currentSpecies={data.species}/>
//         <InnocDisplay innoc={data.innoc} openInNewTab={true}/>
//         <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
//         <ConfirmedCleanArea initial={data.confirmedClean} readonly={true}/>
//         <LcRecipeArea lcRecipeId={data.recipe}/>
//         <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore} offset={-1}/>
//         <NotesAreaInline notes={data.notes} offset={-1}/>
//         <InlineSubArea props={{}}>
//
//
//
//
//
//
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//
//
//         </InlineExpansionArea>
//
//     </InlineEntry>
// }

// export function LcListDisplay({data, onClick}: SingleListProps<LcData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <LcInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }
